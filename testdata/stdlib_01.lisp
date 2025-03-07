(defcolumns A B (C :i16@loob))

(defconstraint c1 () (if-not-eq A B C))
