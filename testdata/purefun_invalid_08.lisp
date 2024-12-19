;;error:6:22-28:expected loobean constraint (found 𝔽)
(defcolumns (X :@loob) Y)
(defpurefun (id x) x)

(defconstraint c1 () (id X))
(defconstraint c2 () (id Y))
