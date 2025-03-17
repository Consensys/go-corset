(defcolumns (A :binary) (B :binary) (C :i16))
(defconstraint c1 () (if (== A B) (== 0 C)))
(defconstraint c2 () (if (== B A) (== 0 C)))
