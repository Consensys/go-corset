;;error:6:22-28:expected loobean constraint (found 𝔽)
(defcolumns (X :@loob) Y)
(defpurefun (fd x) x)

(defconstraint c1 () (fd X))
(defconstraint c2 () (fd Y))
