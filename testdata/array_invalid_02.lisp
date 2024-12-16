;;error:4:24-25:expected loobean constraint (found (𝔽@loob)[4])
(defcolumns (BIT :@loob :array [4]))

(defconstraint bits () BIT)
