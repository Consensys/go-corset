(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16) (B :i16))
(defconstraint c1 () (vanishes! (~ (- A B))))
