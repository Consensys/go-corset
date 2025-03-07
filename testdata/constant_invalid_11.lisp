;;error:4:43-46:not permitted in pure context
(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns X Y TWO)
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
