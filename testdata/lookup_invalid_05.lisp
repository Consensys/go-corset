;;error:2:1-2:blah
(defcolumns X)
(deflookup test (Y) (X))

(module m1)
(defcolumns Y)
