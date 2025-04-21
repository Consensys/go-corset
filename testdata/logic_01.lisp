(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (∨ (== X 0) (== Y 0)))

(defconstraint c2 ()
  (∧ (!= X 1) (!= Y 1)))

(defconstraint c3 ()
  (∧ (¬ (== X 1)) (!= Y 1)))

(defconstraint c4 ()
  (∧ (!= X 1) (¬ (== Y 1))))
