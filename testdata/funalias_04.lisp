(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i48) (B :i48))
(defpurefun (eq x y) (- y x))
(defunalias = eq)
(defconstraint test () (vanishes! (= A B)))
