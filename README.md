App Share
===========

very simple web application proxy for making your local development environment accessible in internet. Good for demostration or debugging purpose.

Usage
------

In your local environment, simply run:

```
$ appshare "manage.py runserver" -p 3000
```

A ``.appshare`` file will be generated in your ``$HOME``. Default server location is ``http://appshare.tchen.me``, for testing purpose only. You could alter this file to put on your own server.


Ideas backing us App Share
----------------------------

When client starts, it will build the proxy channel with server:

```
--> client: PROXY
<-- server: PROXY: blablabla.tchen.me
```