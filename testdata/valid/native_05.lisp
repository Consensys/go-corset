(defcolumns (P :i2) (Q :i2) (X :i16) (Y :i16))
(defcomputed (Z) (fwd-fill-within P Q X))
(defconstraint c1 () (== Y Z))
