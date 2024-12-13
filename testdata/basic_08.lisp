(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y Z)
(defconstraint c1 () (vanishes! (* Z (- Z 1))))
(defconstraint c2 () (vanishes! (* (- Y Z) (- Y Z 2))))
(defconstraint c3 () (vanishes! (* (- X Y) (- X Y 4))))
