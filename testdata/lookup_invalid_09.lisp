;;error:2:1-2:blah
(defcolumns X)
(deflookup test (X) ((+ m2.A 1)))

(module m1)
(defcolumns A)
