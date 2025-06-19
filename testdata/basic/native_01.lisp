(defcolumns (P :i2) (X :i16) (Y :i16))
(defcomputed (Z) (fwd-changes-within P X))
(defconstraint c1 () (== Y Z))
