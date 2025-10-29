(defcolumns (P :i2) (X1 :i16) (X2 :i16) (Y :i16))
(defcomputed (Z) (fwd-changes-within P X1 X2))
(defconstraint c1 () (== Y Z))
