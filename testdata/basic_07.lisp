(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (* Y (- Y 1) (- Y 2) (- Y 3))))
(defconstraint c2 () (vanishes! (* (- X Y) (- X Y 4))))
