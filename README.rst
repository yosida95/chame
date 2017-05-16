chame
=====

chame was inspired a lot by `atmos/camo`_.


Demo
----

A demo instance of chame is deployed and running on Google App Engine at http://chame.yosida95.com/.

Original URL
    https://yosida95.kvs.gehirn.jp/icon_200x200.png
Chame Proxied URL
    http://chame.yosida95.com/i/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2NoYW1lLnlvc2lkYTk1LmNvbSIsInN1YiI6Imh0dHBzOi8veW9zaWRhOTUua3ZzLmdlaGlybi5qcC9pY29uXzIwMHgyMDAucG5nIn0.qWZOZzqKkIz6j6zAsu3DdapYYR8yi32rzxFCA1BvK58


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
