(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns A B)
(defpurefun (eq x y) (- y x))
(defunalias eq! eq)
(defconstraint test () (vanishes! (eq! A B)))
