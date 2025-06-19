;;error:3:26-27:expected bool, found u16
(defcolumns (A :i16) (B :i16) (C :i16))
(defconstraint c1 () (if A (== 0 B) (== 0 C)))
