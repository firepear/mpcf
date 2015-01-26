package main

import(
	//"crypto/md5"
	"database/sql"
	"flag"
	"fmt"
	//"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

var(
	verp = flag.Bool("version", false, "Show version info")
	scanp = flag.Bool("scan", false, "Perform scan of musicdir")
	tagp = flag.Bool("tag", false, "Tag [dir] with [facet]")
	getp = flag.Bool("get", false, "Get filenames for tracks tagged with [facet]")
	musicdir = "/home/mdxi/media/music"
	seen = 0
)

func main() {
	db, err := sql.Open("sqlite3", musicdir + "/.mpdf.db")
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	defer db.Close()

	if *scanp {
		db.Exec("PRAGMA synchronous=0")
		scandir("", db)		
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
	if *verp {
		fmt.Println("This is mpcf v0.1.0")
		os.Exit(0)
	}
	
	// create db if needed
	var tracks int
	res := db.QueryRow("select count(id) from tracks")
	err = res.Scan(&tracks)
	if err != nil {
		log.Println("Creating db")
		createdb(db)
		log.Println("Updating track list")
		db.Exec("PRAGMA synchronous=0")
		scandir("", db)
	} else {
		var tags, facets int
		db.QueryRow("select count(tid) from t2f").Scan(&tags)
		db.QueryRow("select count(id) from facets").Scan(&facets)
		fmt.Printf("%v tracks; %v tagged, with %v facets\n", tracks, tags, facets)
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
	args[0] = strings.TrimLeft(args[0], ".")
	tagdir2(args[0], fid, db)
}

func tagdir2(dir string, fid int, db *sql.DB) {
	err := os.Chdir(musicdir + dir)
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
		"create table tracks (id integer primary key, filename text unique)",
		"create table facets (id integer primary key, facet text)",
		"create table t2f(tid integer, fid integer)",
		"create index fididx on t2f(fid)",
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
			scandir(dir + "/" + direntry.Name(), db)
		} else {
			seen ++
			if seen % 100 == 0 {
				log.Printf("Processed %v tracks\n", seen)
			}
			name := dir + "/" + direntry.Name()
			// if we already have a file with this name, don't do anything else
			var id int
			res := db.QueryRow("select count(id) from tracks where filename = ?", name)
			err = res.Scan(&id)
			if err != nil {
				log.Fatal(err)
			}
			if id > 0 {
				continue
			}
			// nope, this one needs processing
			//md5 := fmt.Sprintf("%x", calcMD5(direntry.Name()))
			//_, err := db.Exec("INSERT OR REPLACE INTO tracks (filename, hash) VALUES(COALESCE((SELECT filename FROM tracks WHERE filename = ?),?), COALESCE((SELECT hash FROM tracks WHERE hash = ?), ?))", name, name, md5, md5)
			_, err := db.Exec("INSERT INTO tracks (filename) VALUES(?)", name)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

/*func calcMD5(filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	return hash.Sum(nil)
}
*/
