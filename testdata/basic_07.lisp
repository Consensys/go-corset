(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (* Y (- Y 1) (- Y 2) (- Y 3))))
(defconstraint c2 () (vanishes! (* (- X Y) (- X Y 4))))
