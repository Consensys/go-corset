(defconst
  ONE 1
  TWO 2)
;;
(module m1)
(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (== 0 (- X (* TWO Y))))
(defconstraint c2 () (== 0 (- (* TWO Y) X)))
(defconstraint c3 () (== 0 (- X Y Y)))
(defconstraint c6 () (== 0 (- (* ONE X) Y Y)))
