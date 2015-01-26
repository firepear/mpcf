**********************************
mpcf
**********************************
Faceted classification for mpc/mpd
==================================

If you use mpd/mpc to listen to your music, mpcf will let you add
facets (cross-cutting classifications) to your music collection. For
instance, I have facets for locations (Atlanta bands, etc.), for music
origin (things I ripped from vinyl), and so on (bands I've seen
live). Views like this do not fit into existing ID3 implementations.

Tracks can, of course, be tagged with arbitrary numbers of facets.

It is assumed that your music is organized by directory in some useful
manner; tagging tracks operates on whole directory subtrees.

::
   
    # initialize db, or get a quick status report
    $ mpcf
    # tag files in subtree DIRECTORY with FACET
    $ mpcf -tag [DIRECTORY] [FACET]
    # show extant facets
    $ mpcf -facets
    # add all files with FACET to playlist
    $ mpcf -get [FACET] | mpc add
    # scan musicdir for new/modified files
    $ mpcf -scan

Applying tags is idempotent, so don't worry about whether something
has or has not already been tagged.

If a file's location changes _or_ if the file itself changes but is in
the same location, then tags will stick with the file. If a file moves
*and* changes, however, it will be seen as new (have no tags
associated with it), requiring a sync.

* Current version: 0.3.0 (2015-01-25)

* `Issue tracker <https://firepear.atlassian.net/browse/MPCF>`_

* Source repo: :code:`git://firepear.net/goutils/mcpf.git`


Send questions, suggestions, or problem reports to shawn@firepear.net
