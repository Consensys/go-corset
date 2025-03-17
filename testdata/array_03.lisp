(defcolumns (ARR :i16 :array [2]))
(defconstraint c1 () (== [ARR 1] [ARR 2]))
(defconstraint c2 () (== [ARR 2] [ARR 1]))
