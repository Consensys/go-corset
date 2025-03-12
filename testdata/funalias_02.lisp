(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i32) (B :i32))
(defpurefun (eq x y) (- y x))
(defunalias eq! eq)
(defconstraint test () (vanishes! (eq! A B)))
