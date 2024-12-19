;;error:4:18-19:conflicting context
(defcolumns X Y)
(definterleaved A (X Y))
(deflookup l1 (A X) (Y Y))
