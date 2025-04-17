(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (∨ (== X 0) (== Y 0)))

(defconstraint c2 ()
  (∧ (!= X 1) (!= Y 1)))
