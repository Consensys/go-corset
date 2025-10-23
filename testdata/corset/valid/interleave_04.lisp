(defcolumns (X :i16) (Y :i16) (Z :i16))
(definterleaved A (X Y))
(deflookup l1 (A) (Z))
