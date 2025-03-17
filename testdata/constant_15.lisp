(defpurefun (vanishes! x) (== 0 x))

(defconst
  (ONE :extern)   0x01
  (TWO :extern)   0x02
  THREE 0x03
  FOUR  0x04
)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (* Y (- Y ONE) (- Y TWO) (- Y THREE))))
(defconstraint c2 () (vanishes! (* (- X Y) (- X Y FOUR))))
