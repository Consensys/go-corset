(defconst
  ONE_  1
  ONE   ONE_
  TWO   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 () (== 0 (* Z (- Z ONE))))
(defconstraint c2 () (== 0 (* (- Y Z) (- Y Z TWO))))
(defconstraint c3 () (== 0 (* (- X Y) (- X Y FOUR))))
