gobmp
=====

A Go package for reading and writing BMP image files.


Installation
------------

To download and install, at a command prompt type:

    go get github.com/jsummers/gobmp


Documentation
-------------

Gobmp is designed to work the same as Go's standard
[image modules](http://golang.org/pkg/image/).

The (minimal) documentation may be read online at
[GoPkgDoc](http://go.pkgdoc.org/github.com/jsummers/gobmp).

Or, after installing, type:

    godoc github.com/jsummers/gobmp | more


Status
------

The encoder writes all images as 24-bit truecolor images. I plan to update
it to write paletted images in some cases.

The decoder supports most uncompressed BMP images. I plan to update it to
support nearly all BMP image formats.


License
-------

Gobmp is distributed under an MIT-style license.

Copyright &copy; 2012 Jason Summers
<[jason1@pobox.com](mailto:jason1@pobox.com)>

<pre>
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
</pre>
