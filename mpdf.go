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
	_ "github.com/mattn/go-sqlite3"
)

var(
	scanp = flag.Bool("scan", false, "Perform scan of musicdir")
	tagp = flag.Bool("tag", false, "Tag [dir] with [facet]")
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
		tagdir(flag.Args())
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
		db.QueryRow("select count(tid) from t2f").Scan(tags)
		db.QueryRow("select count(id) from facets").Scan(facets)
		fmt.Printf("%v tracks; %v tagged, with %v facets\n", tracks, tags, facets)
	}
}

func tagdir([]string) {
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
