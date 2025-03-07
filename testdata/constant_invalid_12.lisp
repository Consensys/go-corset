;;error:6:32-37:not permitted in pure context
;;error:6:22-39:expected loobean constraint (found 𝔽)
(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (X :i16) (Y :i16) (Z :i16))
(defun (TWO) Z)
(defconstraint c1 () (- Y (^ X (TWO))))
