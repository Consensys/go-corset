;;error:7:43-46:not permitted in const context
(defpurefun ((vanishes! :𝔽@loob) x) x)

(defconst (TWO :extern) 2)
(defcolumns (X :i16) (Y :i16))
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
