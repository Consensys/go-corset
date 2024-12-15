;;error:2:1-2:blah
(defcolumns X)
(deflookup test (X) (Y))

(module m1)
(defcolumns Y)
