(defcolumns (X :i16@loob) (Y :i16@loob))
(definterleaved A (X Y))
(definterleaved B (X Y))
(definterleaved Z (A B))
(defconstraint c1 () Z)
