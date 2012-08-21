`seamless` is a TCP proxy that allow you to deploy new code then switch traffic
to new backend without downtime.

Switching backends is done with HTTP interface (*on a different port*) with the
following API:

    `/switch?backend=address` 
        switch traffic to new backend

    `/current` 
        return (in plain text) current server

Process
=======
* Start first backend at port 4444
* Run
  ::

    seamless 8080 localhost:4444
* Direct all traffic to port 8080 on local machine.

When you need to upgrade the backend, start a new one (with new code on a
different port, say 4445). Then::

    curl http://localhost:6777/switch?backend=localhost:4445


(Note that management port is different from the one we proxy).

Installing
==========
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
