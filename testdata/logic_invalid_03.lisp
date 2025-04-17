;;error:5:14-26:expected int, found bool
;;error:7:14-28:expected int, found bool
(defcolumns (X :i16))
(defconstraint c1 ()
  (== X (+ 1 (∨ (== X 0)))))
(defconstraint c2 ()
  (== X (* 2 (∧ 1 (== X 0)))))
