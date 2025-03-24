;;error:9:13-21:symbol X already declared
(defconst
  X     1
  ONE   X
  TWO   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 () (== 0 (* Z (- Z ONE))))
(defconstraint c2 () (== 0 (* (- Y Z) (- Y Z TWO))))
(defconstraint c3 () (== 0 (* (- X Y) (- X Y FOUR))))
