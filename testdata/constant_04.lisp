(defpurefun (vanishes! x) (== 0 x))

(defconst
  ONE_  1
  ONE   ONE_
  TWO   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 () (vanishes! (* Z (- Z ONE))))
(defconstraint c2 () (vanishes! (* (- Y Z) (- Y Z TWO))))
(defconstraint c3 () (vanishes! (* (- X Y) (- X Y FOUR))))
