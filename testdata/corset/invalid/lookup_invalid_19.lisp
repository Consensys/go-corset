(module m1)
(defcolumns (X :i16) (Y :i16))
(defclookup test (Y) 1 (X))

(module m1)
(defclookup test (X) 1 (Y))
