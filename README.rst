Geolocation
=======================
This is a simple geolocation service.

It provides the remote ip (via X-Forwarded-For header, if present), country code, city, and geographical coordinates.
Information is provided in plain text format, under ``/``, and in json, under ``/json``.

Prerequisites
-----------------------

You can use ``geoipupdate`` to download MaxMind's City database::

    sudo apt install geoipupdate
    sudo cp /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
    sudo geoipupdate -v

Usage
-----------------------

-geodb <path>
	path to the GeoLite2-City database (default is "/var/lib/GeoIP/GeoLite2-City.mmdb")
-port <port>
	port where the service listens on (default is 9001)

