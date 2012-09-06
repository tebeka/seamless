`seameless` is a TCP proxy that allow you to deploy new code then switch traffic
to it without downtime.

It does "round robin" between the list of current active backends.

Switching server is done with HTTP interface with the following API:

/set?backends=host:port,host:port
    set list of backends

/add?backend=host:port
    add a backend

/remove?backend=host:port
    remove a backend

/get
    return host:port,host:port

Process
=======
* Start `seamleass` with list of active backends::

    seamless 8080 localhost:4444
* Direct all traffic to port 8080 on local machine.
* When you need to add/remove the backend, use the HTTP API on port 6777
  different port, say 4445)::

    curl http://localhost:6777/add?backend=localhost:4445
    curl http://localhost:6777/remove?backend=localhost:4444

  Or::

        curl http://localhost:6777/set?backends=localhost:4445
    
New traffic will be directed to new backend(s).

Installing
==========
You can download a statically linked executable at the downloads_ section.

.. _downloads: https://bitbucket.org/tebeka/seamless/downloads

Or if you have a Go development environment, you can

::

    go get bitbucket.org/tebeka/seamless

Contact
=======
Miki Tebeka <miki.tebeka@gmail.com> or here_.

.. _here: https://bitbucket.org/tebeka/seamless


LICENSE
=======
MIT_

.. _MIT: https://bitbucket.org/tebeka/seamless/src/tip/LICENSE.txt
