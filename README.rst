**********************************
mpcf
**********************************
Faceted classification for mpc/mpd
==================================

If you use mpd/mpc to listen to your music, mpcf will let you add
facets (cross-cutting classifications) to your music collection.

::
   
    # initialize db, or get a quick status report
    $ mpcf
    # tag files in subtree DIRECTORY with FACET
    $ mpcf -tag [DIRECTORY] [FACET]
    # add all files with FACET to playlist
    $ mpcf -get [FACET] | mpc add
    # scan musicdir for new/modified files
    $ mpcf -scan

It's not ready for use by other people yet. This is informational.

* Current version: 0.3.0 (2015-01-25)

* `Issue tracker <https://firepear.atlassian.net/browse/MCPF>`_

* Source repo: :code:`git://firepear.net/goutils/mcpf.git`


Send questions, suggestions, or problem reports to shawn@firepear.net
