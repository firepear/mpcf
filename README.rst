**********************************
mpcf
**********************************
Faceted classification for mpc/mpd
==================================

*mpcf is unmaintained software*

If you use mpd/mpc to listen to your music, mpcf will let you add
facets — cross-cutting classifications — to your music collection. For
instance, I have facets for locations (Atlanta bands), for music
origin (things I ripped from vinyl), and so on (bands I've seen live).

Tracks can, of course, be tagged with arbitrary numbers of facets.

It is assumed that your music is organized by directory in some useful
manner (*e.g.* :code:`musicdir/artist/album`) as tagging of tracks
operates on directory subtrees.

::
   
    # first run; set music directory and initialize db
    $ mpcf -musicdir=/path/to/mpd/musicdir
    
    # tag all files in subtree DIRECTORY with FACET
    $ mpcf -tag [DIRECTORY] [FACET]
    
    # add all files with FACET to playlist
    $ mpcf -get [FACET] | mpc add
    
    # update and sweep mcpf db
    $ mpcf -scan

See :code:`'mcpf -h'` for (slightly) more capabilities.
    
Applying tags is idempotent, so don't worry about whether something
has or has not already been tagged.

If a file's location changes *or* if the file itself changes but is in
the same location, then tags will stick with the file when
:code:`'mcpf -scan'` is run. However, if a file moves *and* changes,
it will be seen as new and no longer have tags associated with it.

mpcf's database is stored at :code:`MUSICDIR/.mpcf.db`, so if you
backup your music directory, the db should be backed up as well.

There are plans to add (slightly) richer functionality and more
polish, but the db format is likely stable.

* Current version: 0.5.3 (2015-04-28)

* Install: :code:`go get firepear.net/mpcf`

* `Github <https://github.com/firepear/mpcf/>`_
