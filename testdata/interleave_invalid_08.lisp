;;error:2:1-2:blah
(defcolumns X Y)
(definterleaved A (X Y))
(deflookup l1 (A X) (Y Y))
