;;error:4:24-27:expected loobean constraint (found (𝔽@loob)[1:4])
(defcolumns (BIT :@loob :array [4]))

(defconstraint bits () BIT)
