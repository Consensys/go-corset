(defcolumns (A :i32) (B :i16) (C :i32@loob))

(defconstraint c1 () (if-not-eq A B C))
