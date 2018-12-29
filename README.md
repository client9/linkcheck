# linkcheck
Checks links in static websites before being published

## What Problem Does It Solve?

Making sure internal links (links pointing to other documents in the same site) are working.

While there are a few other services that test links on the live site, it's even better to find problems before they are published.

Likewise, most static website generators have a method to do cross-referencing (e.g. [Hugo](https://gohugo.io/content-management/cross-references/)) but it's sometimes clunky and it's not enforced.  Most of the time that only works for content.  Another problem is that links can be in the template, and be generated as well.

LinkCheck works off the output of the static site generator and checks all cross-references.

## Status

Embarrassing alpha, but works.

## Todo

* Clean up code so other people can work on this
* Check linking to non-HTML internal resources
* Remove hostname from fully specified internal links
* Check for URL normalization
* Check for external dead links
* Allow or warn on links that are not HTTPS
* Checking URL fragments match anchors
* Find links to aliases that should the primary permalink (See [Hugo and Aliases](https://gohugo.io/content-management/urls/#aliases)).
