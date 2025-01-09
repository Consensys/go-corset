(defpurefun ((vanishes! :@loob) x) x)
(defpurefun (f x) (let ((xp1 (+ x 1))) xp1))

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (if A
      (vanishes! (f (- B 1)))))
