(defcolumns (X :i16) (Y :i16))
(definterleaved A (X Y))
(definterleaved B (X Y))
(definterleaved Z (A B))
(defconstraint c1 () (== 0 Z))
