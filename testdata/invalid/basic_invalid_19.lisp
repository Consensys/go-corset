;;error:6:16-18:duplicate handle
(defconst (C :extern :i64) 1)
(defcolumns (X :i64))

(defconstraint c1 () (== X C))
(defconstraint c1 () (== X C))
