(defcolumns A B (C :@loob))

(defconstraint c1 ()
  (if-not-eq A B C))
