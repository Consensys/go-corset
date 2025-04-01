(defcolumns (X :i16) (Y :i16))
(definterleaved Z (X Y))
(defconstraint c1 () (== 0 Z))
