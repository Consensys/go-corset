(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16))
;; X*2 == Y*2
(defconstraint c1 () (vanishes! (- (* X (* 2 1)) (* Y (* 1 2)))))
;; X*1458 == Y*1458
(defconstraint c2 () (vanishes! (- (* X (* 243 22)) (* Y (* 6 891)))))
;; X*2916 == Y*2916
(defconstraint c3 () (vanishes! (- (* X (* 2 243 22)) (* Y (* 6 891 2)))))
;; X*2916 == Y*2916
(defconstraint c4 () (vanishes! (- (* X (* 243 2 22)) (* Y (* 6 891 2)))))
;; X*2916 == Y*2916
(defconstraint c5 () (vanishes! (- (* X (* 22 243 2)) (* Y (* 6 891 2)))))
;; X*2916 == Y*2916
(defconstraint c6 () (vanishes! (- (* X (* 2 243 22)) (* Y (* 891 6 2)))))
;; X*2916 == Y*2916
(defconstraint c7 () (vanishes! (- (* X (* 2 243 22)) (* Y (* 2 891 6)))))
;; X*2916 == Y*2916
(defconstraint c8 () (vanishes! (- (* X (* 2 243 22)) (* Y (* 2 891 6 1)))))
;; X*2916 == Y*2916
(defconstraint c9 () (vanishes! (- (* X (* 2 243 22)) (* Y (* 2 891 6 1 1)))))
;; X*2916 == Y*2916
(defconstraint c10 () (vanishes! (- (* X (* 2 243 22 1)) (* Y (* 2 891 6)))))
;; X*2916 == Y*2916
(defconstraint c11 () (vanishes! (- (* X (* 2 243 22 1 1)) (* Y (* 2 891 6)))))
