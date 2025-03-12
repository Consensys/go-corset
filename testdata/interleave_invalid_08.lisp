;;error:4:18-19:conflicting context
(defcolumns (X :i16) (Y :i16))
(definterleaved A (X Y))
(deflookup l1 (A X) (Y Y))
