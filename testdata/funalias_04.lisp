(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns A B)
(defpurefun (eq x y) (- y x))
(defunalias = eq)
(defconstraint test () (vanishes! (= A B)))
