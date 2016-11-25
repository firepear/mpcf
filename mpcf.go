package main

import (
	"crypto/md5"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	verp     = flag.Bool("version", false, "Show version info")
	facetp   = flag.Bool("facets", false, "List facets")
	scanp    = flag.Bool("scan", false, "Check for new/modified files and sweep db for orphan records")
	cleanp   = flag.Bool("cleandb", false, "Just clean db (no file scan)")
	tagp     = flag.Bool("tag", false, "Tag [dir] with [facet]")
	getp     = flag.Bool("get", false, "Get filenames for tracks tagged with [facet]")
	mdflag   = flag.String("musicdir", "", "Set location of your mpd music directory")
	musicdir = ""
	seen     int64
	touched  int64
)

func init() {
	flag.Parse()
	config := os.Getenv("HOME") + "/.mpcf"
	if *mdflag != "" {
		err := ioutil.WriteFile(config, []byte(*mdflag), 0644)
		if err != nil {
			log.Fatal(err)
		}
		musicdir = *mdflag
	} else {
		mdbytes, err := ioutil.ReadFile(config)
		if err != nil {
			log.Fatal("Please run 'mpcf -musicdir /path/to/music' to set your musicdir path.")
		}
		musicdir = string(mdbytes)
	}
}

func main() {
	db, err := sql.Open("sqlite3", musicdir+"/.mpcf.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// create db if needed
	var tracks int
	res := db.QueryRow("select count(id) from tracks")
	err = res.Scan(&tracks)
	if err != nil {
		db.Exec("PRAGMA synchronous=0")
		log.Println("Creating db")
		createdb(db)
		log.Println("Updating track list")
		scandir("", db)
	}

	if *verp {
		fmt.Println("This is mpcf v0.5.3")
		os.Exit(0)
	}
	if *scanp {
		db.Exec("PRAGMA synchronous=0")
		scandir("", db)
		cleandb(db)
		os.Exit(0)
	}
	if *cleanp {
		cleandb(db)
		os.Exit(0)
	}
	if *tagp {
		tagdir(flag.Args(), db)
		os.Exit(0)
	}
	if *getp {
		getfacettracks(flag.Args(), db)
		os.Exit(0)
	}
	if *facetp {
		lsfacets(db)
		os.Exit(0)
	}

	var taggedtracks, tags, facets int
	db.QueryRow("select count(tid) from t2f").Scan(&tags)
	db.QueryRow("select count(distinct tid) from t2f").Scan(&taggedtracks)
	db.QueryRow("select count(id) from facets").Scan(&facets)
	fmt.Printf("%v tracks (%v tagged)\n%v tags\n%v facets\n", tracks, taggedtracks, tags, facets)
}

func lsfacets(db *sql.DB) {
	rows, err := db.Query("SELECT facet FROM facets ORDER BY facet")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var f string
	for rows.Next() {
		if err := rows.Scan(&f); err != nil {
			log.Fatal(err)
		}
		fmt.Println(f)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func getfacettracks(args []string, db *sql.DB) {
	if len(args) != 1 {
		log.Fatal("Too many/few arguments to -get; need a facet name")
	}
	var fid int
	db.QueryRow("select id from facets where facet = ?", args[0]).Scan(&fid)
	if fid == 0 {
		return
	}
	rows, err := db.Query("SELECT filename FROM tracks WHERE id IN (SELECT DISTINCT tid FROM t2f WHERE fid = ?)", fid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		fmt.Println(name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func tagdir(args []string, db *sql.DB) {
	if len(args) != 2 {
		log.Fatal("Too many/few arguments to -tag; need a directory and a facet")
	}
	// create the tag if it doesn't exist
	var fid int
	db.QueryRow("select id from facets where facet = ?", args[1]).Scan(&fid)
	if fid == 0 {
		db.Exec("insert into facets (facet) values (?)", args[1])
		db.QueryRow("select id from facets where facet = ?", args[1]).Scan(&fid)
	}
	// now actually tag tracks under this dir
	args[0] = strings.TrimRight(args[0], "/")
	args[0] = strings.TrimLeft(args[0], "./")
	args[0] = strings.TrimLeft(args[0], musicdir)
	tagdir2(args[0], fid, db)
}

func tagdir2(dir string, fid int, db *sql.DB) {
	err := os.Chdir(musicdir + "/" + dir)
	if err != nil {
		log.Fatalf("Can't chdir to %v", dir)
	}
	ls, err := ioutil.ReadDir(".")
	for _, direntry := range ls {
		name := dir + "/" + direntry.Name()
		if direntry.IsDir() {
			tagdir2(name, fid, db)
		} else {
			var tid, fcnt int
			db.QueryRow("select id from tracks where filename = ?", name).Scan(&tid)
			db.QueryRow("select count(tid) from t2f where tid = ? and fid = ?", tid, fid).Scan(&fcnt)
			if fcnt > 0 {
				continue
			}
			db.Exec("insert into t2f (tid, fid) values (?, ?)", tid, fid)
		}
	}
}

func createdb(db *sql.DB) {
	var err error
	var stmts = []string{
		"create table tracks (id integer primary key, filename text unique, hash text unique)",
		"create table facets (id integer primary key, facet text)",
		"create table t2f (tid integer, fid integer)",
		"create index fididx on t2f(fid)",
		"create table config (key text, value text)",
		"insert into config (key, value) values('mpdconf', '/etc/mpd.conf')",
	}
	for _, stmt := range stmts {
		if err != nil {
			break
		}
		_, err = db.Exec(stmt)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func scandir(dir string, db *sql.DB) {
	os.Chdir(musicdir + "/" + dir)
	ls, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err, dir)
	}
	for _, direntry := range ls {
		if direntry.IsDir() {
			if dir == "" {
				scandir(direntry.Name(), db)
			} else {
				scandir(dir+"/"+direntry.Name(), db)
			}
		} else {
			seen++
			if seen%100 == 0 {
				log.Printf("Processed %v tracks; updated %v\n", seen, touched)
			}
			name := dir + "/" + direntry.Name()
			md5 := fmt.Sprintf("%x", calcMD5(direntry.Name()))
			// _, err := db.Exec("INSERT OR REPLACE INTO tracks (filename, hash) VALUES(COALESCE((SELECT filename FROM tracks WHERE filename = ?),?), COALESCE((SELECT hash FROM tracks WHERE hash = ?), ?))", name, name, md5, md5)
			r, err := db.Exec("INSERT OR IGNORE INTO tracks (filename, hash) VALUES(?, ?)", name, md5)
			if err != nil {
				log.Fatal(err)
			}
			touch, _ := r.RowsAffected()
			touched += touch
			//r, err = db.Exec("UPDATE tracks SET filename = ?, hash = ? WHERE filename = ?", name, md5, name)
			//if err != nil {
			//	log.Fatal(err)
			//}
		}
	}
}

func cleandb(db *sql.DB) {
	log.Printf("Scanning db for orphaned records")
	rows, err := db.Query("SELECT id, filename FROM tracks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var id int64
	var name string
	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		_, err = os.Stat(musicdir + "/" + name)
		if err == nil {
			continue
		}
		// remove track entry
		_, err = db.Exec("delete from tracks where id = ?", id)
		if err != nil {
			log.Fatal(err)
		}
		// remove tag links
		_, err = db.Exec("delete from t2f where tid = ?", id)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Removed orphan record for %v\n", name)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("vacuum")
}

func calcMD5(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.CopyN(hash, file, 524288); err != nil && err != io.EOF {
		log.Fatal(err)
	}
	return hash.Sum(nil)
}
