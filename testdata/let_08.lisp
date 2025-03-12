(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (A :i16@loob) (B :i16))
(defconstraint c1 ()
  (let ((B (+ B 1)))
    (if A
        (vanishes! (- B 1)))))
