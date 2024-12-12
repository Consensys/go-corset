(defpurefun ((vanishes! :@loob) x) x)

(defconst
  ONE   0x01
  TWO   0x02
  THREE 0x03
  FOUR  0x04
)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (* Y (- Y ONE) (- Y TWO) (- Y THREE))))
(defconstraint c2 () (vanishes! (* (- X Y) (- X Y FOUR))))
