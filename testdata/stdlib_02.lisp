(defcolumns (A :i16) (B :i16) (C :i16@loob))

(defconstraint c1 ()
  (if-not-eq A B C))
