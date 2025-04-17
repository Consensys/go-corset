;;error:5:15-16:expected bool, found u1
;;error:7:6-7:expected bool, found u1
(defcolumns (X :i16))
(defconstraint c1 ()
  (∨ (== X 0) 1))
(defconstraint c2 ()
  (∧ 1 (== X 0)))
