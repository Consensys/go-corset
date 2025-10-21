(defcolumns (X :binary) (Y :binary) (A :i16))
(defconstraint c1 () (if (== 0 X) (== 0 A)))
(defconstraint c2 () (if (!= 0 Y) (== 0 A)))
