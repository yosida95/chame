chame
=====

chame was inspired a lot by `atmos/camo`_.


Demo
----

A demo instance of chame is deployed and running on Google App Engine at https://chame.yosida95.com/.

Original URL
    http://kvs.gehirn.jp/yosida95/icon_200x200.png
Chame Proxied URL
    https://chame.yosida95.com/i/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2NoYW1lLnlvc2lkYTk1LmNvbSIsInN1YiI6Imh0dHA6Ly9rdnMuZ2VoaXJuLmpwL3lvc2lkYTk1L2ljb25fMjAweDIwMC5wbmcifQ.2LztWUS-DMv5mNSsdYQHsEc7tiXUz1YIvALh4fcvcek


Quick Start
-----------

Download and Install
~~~~~~~~~~~~~~~~~~~~

.. code-block:: console

    $ go get -u -v github.com/yosida95/chame


Start chame server
~~~~~~~~~~~~~~~~~~

.. code-block:: console

    $ $GOBIN/chame serve


Documentation
-------------

.. code-block:: console

    $ $GOBIN/chame
    Usage:
      chame [command]

    Available Commands:
      encode
      decode
      serve

    Use "chame [command] --help" for more information about a command.


Package documentation is available on GoDoc.

pkg/chame
    https://godoc.org/github.com/yosida95/chame/pkg/chame
pkg/metadata
    https://godoc.org/github.com/yosida95/chame/pkg/metadata
pkg/memstore
    https://godoc.org/github.com/yosida95/chame/pkg/memstore


Deploy to Google App Engine
---------------------------

chame is designed to intend to be deployed easily on `Google App Engine`_.
You can deploy chame with Google's `Cloud SDK`_ like below.

.. code-block:: console

    $ cd $GOPATH/src/github.com/yosida95/chame/appengine
    $ gcloud app deploy --project YOUR_PROJECT app.yaml


User(s)
-------

- `Gehirn Inc. <https://www.gehirn.co.jp/>`_ ( https://chame.usercontent.jp/, My employer )


Author
------

Kohei YOSHIDA a.k.a. yosida95_


License
-------

chame is distributed under the Apache License Version 2.0.
See ./LICENSE.

.. _yosida95: https://yosida95.com/
.. _`atmos/camo`: https://github.com/atmos/camo

.. _`Google App Engine`: https://cloud.google.com/appengine/
.. _`Cloud SDK`: https://cloud.google.com/sdk/
