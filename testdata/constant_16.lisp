(defpurefun ((vanishes! :@loob) x) x)

(defconst
  (ONE_ :extern)  1
  ONE   ONE_
  (TWO :extern)   (+ 1 ONE)
  FOUR  (* 2 TWO)
)

(defcolumns X Y Z)
(defconstraint c1 () (vanishes! (* Z (- Z ONE))))
(defconstraint c2 () (vanishes! (* (- Y Z) (- Y Z TWO))))
(defconstraint c3 () (vanishes! (* (- X Y) (- X Y FOUR))))
