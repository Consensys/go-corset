(defcolumns (P :binary) (K1a :i32) (K1b :i32) (V1 :i16) (K2a :i32) (K2b :i32) (R :i16))
(defcomputed (V2) (map-if P K2a K2b P K1a K1b V1))
(defconstraint c1 () (== V2 R))
