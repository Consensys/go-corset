(defcolumns (X :i16))

(defconstraint c1 ()
  (∧ (== 0 1) (!= X 1)))
