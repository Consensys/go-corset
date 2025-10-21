(defcolumns (X :i16) (Y :i16))
(definterleaved (Z :i16) (X Y))
(defconstraint c1 () (== 0 Z))
