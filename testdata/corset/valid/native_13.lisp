(defcolumns (P :binary) (Q :binary) (X :i16) (Y :i16))
(defcomputed (Z) (bwd-fill-within P Q X))
(defconstraint c1 () (== Y Z))
