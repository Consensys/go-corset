;;error:1:1-1:impure
(defpurefun ((vanishes! :@loob) x) x)
(defpurefun (f x) (let ((xp1 (+ A 1))) xp1))

(defcolumns (A :@loob) B)
(defconstraint c1 ()
  (if A
      (vanishes! (f (- B 1)))))
