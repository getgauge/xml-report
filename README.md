xml-report
==========

 [ ![Download Nightly](https://api.bintray.com/packages/gauge/xml-report/Nightly/images/download.svg) ](https://bintray.com/gauge/xml-report/Nightly/_latestVersion) [![Build Status](https://travis-ci.org/getgauge/xml-report.svg?branch=master)](https://travis-ci.org/getgauge/xml-report)
 [![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v1.4%20adopted-ff69b4.svg)](CODE_OF_CONDUCT.md)

XML Report plugin creates JUnit XML test result document that can be
read by tools such as Go, Jenkins. The format of
XML report is based on [JUnit XML Schema](https://windyroad.com.au/dl/Open%20Source/JUnit.xsd).

**Sample XML Report Document** :

```xml
    <testsuites>
        <testsuite id="1" tests="1" failures="0" package="specs/hello_world.spec" time="0.002" timestamp="2015-09-09T13:52:00" name="Specification Heading" errors="0" hostname="INcomputer.local">
            <properties></properties>
            <testcase classname="Specification Heading" name="First scenario" time="0.001"></testcase>
            <system-out></system-out>
            <system-err></system-err>
        </testsuite>
    </testsuites>
```


Installation
------------

````
gauge install xml-report
````

* Installing specific version

```
gauge install xml-report --version $VERSION
```

### Offline installation

* Download the plugin from [Releases](https://github.com/getgauge/xml-report/releases)
```
gauge install xml-report --file xml-report-$VERSION-$OS.$ARCH.zip
```

Configuration
------------

To add XML report plugin to your project, run the following command :

```
gauge install xml-report
```

The XML report plugin can be configured by the properties set in the
``env/default.properties`` file in the project.

The configurable properties are:

**gauge_reports_dir**

Specifies the path to the directory where the execution reports will be generated.

-  Should be either relative to the project directory or an absolute
   path. By default it is set to `reports` directory in the project.

**overwrite_reports**

Set to `true` if the reports **must be overwritten** on each execution hence maintaining only the latest
execution report.

-  If set to `false` then a **new report** will be generated on each
   execution in the reports directory in a nested time-stamped
   directory. By default it is set to `true`.


License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
Xml-Report is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2018 ThoughtWorks, Inc.


