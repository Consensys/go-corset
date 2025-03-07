(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns X)
(module m1)
;; Module without any column declarations to test alignment.
(module m2)
(defcolumns X)
(defconstraint heartbeat () (vanishes! X))
