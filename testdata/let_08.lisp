(defpurefun ((vanishes! :@loob) x) x)

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (let ((B (+ B 1)))
    (if A
        (vanishes! (- B 1)))))
